# ML Model Documentation and Report

## Abstract

The model present in [NetSlicer1.0.ipynb](https://colab.research.google.com/drive/1kIW3iXNCFchANyA6SV9NESXJDxVTRcmG?usp=sharing#scrollTo=aH5Ts6ns0Ucx) is a machine learning solution for accurately assigning devices to the correct network slice type (eMBB, URLLC, massive IoT) in modern 5G networks. The system uses real telemetry data including latency, packet loss, and service flags to classify devices into slice categories 1, 2, or 3, ensuring optimal QoS and resource efficiency.

Rather than relying on hard-coded thresholds, NetSlicer employs a data-driven Random Forest classifier that can gracefully handle uncertain or missing measurements.

## Section 1: Introduction & Data Loading

The system begins by loading `train_dataset.csv` into a Pandas DataFrame containing 16 raw columns:
- **4 continuous features**: LTE/5G Category, Time, Packet Loss Rate, Packet Delay
- **12 boolean flags**: IoT, LTE/5G, GBR, and one-hot encoded service types

Initial data exploration confirms:
- Continuous fields span meaningful ranges (e.g., packet delay 10–300 ms)
- Boolean features are mostly 0/1 with few missing entries
- Class distribution is roughly 50%, 22%, and 22% for slice types 1, 2, and 3 respectively

A stratified 80/20 train/test split preserves class balance during evaluation.

## Section 2: Exploratory Data Analysis

Correlation analysis reveals:
- Packet delay and Packet Loss Rate show moderate positive correlation (~0.45)
- Continuous features remain largely uncorrelated with delay/loss
- Boolean flags show near-zero pairwise correlation, confirming distinct service indicators

## Section 3: Preprocessing & Feature Reduction

Preliminary experiments identified that "IoT" and "Non-GBR" flags provided no additional signal beyond other service indicators. These redundant columns were removed to:
- Simplify the model
- Reduce overfitting risk
- Eliminate collinear features

Final feature set: 4 continuous columns + 10 boolean indicators

## Section 4: Baseline Random Forest

Initial Random Forest (100 trees) achieved perfect accuracy (1.00) and zero log-loss on the test set. However, such flawless performance typically indicates data leakage or memorization rather than genuine generalization.

## Section 5: Baseline Feature Importances

Feature importance analysis revealed:
- Boolean flags dominated the top ranks
- Continuous metrics (packet delay, packet loss) ranked much lower
- Model relied primarily on service-type flags rather than actual network quality

This unbalanced reliance suggested vulnerability to missing boolean data.

## Section 6: Stress-Testing with Missingness Injection

To verify the model's over-reliance on boolean flags, we simulated dropouts in continuous features at 25%, 50%, and 75% missingness levels for:
- Packet Loss Rate only
- Packet Delay only  
- Both simultaneously

**Result**: The model maintained perfect performance despite massive gaps in continuous data, confirming it never truly learned from network metrics.

## Section 7: Quantifying Leakage with Permutation Importance

Permutation importance analysis revealed:
- "LTE/5G" flag had overwhelming importance (~15% accuracy drop when shuffled)
- All continuous metrics showed near-zero importance
- Model essentially learned to equate service flags with slice type

**Key Finding**: The model exhibited definitive data leakage by using boolean flags as perfect proxies for slice classification.

## Section 8: Overcoming Leakage via Three-State Boolean Encoding

To address the leakage issue, we implemented a three-state encoding for boolean features:
- `1`: True
- `0`: False  
- `-1`: Unknown/Missing

This encoding prevents the model from using missing flag patterns as implicit signals, forcing it to learn from actual network performance metrics when service flags are unavailable.

## Section 9: Robust Model Evaluation & Calibration

With leak-proof encoding implemented, we re-evaluated performance under varying boolean missingness rates:

| Missing Rate | Accuracy | Log Loss | Brier Score | OOB Error |
|--------------|----------|----------|-------------|-----------|
| 0%           | 0.97     | 0.05     | 0.04        | 0.03      |
| 25%          | 0.95     | 0.08     | 0.06        | 0.05      |
| 50%          | 0.90     | 0.15     | 0.10        | 0.10      |
| 75%          | 0.75     | 0.40     | 0.25        | 0.30      |

**Result**: Performance now degrades gracefully as boolean information becomes unavailable, validating that the three-state encoding forces the model to leverage genuine network features.

## Final Implementation

The complete evaluation pipeline includes:
- Mean imputation for missing values
- Random Forest with 200 trees and OOB scoring
- Comprehensive probabilistic metrics (log-loss, Brier score)
- Calibration curve analysis

The final model was exported as an ONNX binary for production deployment.

## Conclusion

NetSlicer 1.0 successfully demonstrates how to build a robust network slice classifier that:
1. Avoids data leakage through careful feature encoding
2. Gracefully handles missing service information
3. Learns meaningful patterns from actual network performance metrics
4. Maintains good calibration across different missing data scenarios

This approach ensures reliable 5G network slice assignment even under uncertain or incomplete telemetry conditions.