(module add)

(defcolumns
  (ACC_2 :i128)
  (RES_LO :i128)
  (ARG_1_LO :i128)
  (OVERFLOW :binary@prove)
  (RES_HI :i128)
  (INST :i8)
  (BYTE_1 :byte@prove)
  (BYTE_2 :byte@prove)
  (ACC_1 :i128)
  (STAMP :i32)
  (ARG_1_HI :i128)
  (ARG_2_LO :i128)
  (ARG_2_HI :i128)
  (CT_MAX :i8)
  (CT :i8))

(defconstraint adder-constraints () (if STAMP 0 (if (- CT CT_MAX) (begin (- RES_HI ACC_1) (- RES_LO ACC_2) (if (- INST 3) 0 (begin (- (+ ARG_1_LO ARG_2_LO) (+ RES_LO (* 340282366920938463463374607431768211456 OVERFLOW))) (- (+ ARG_1_HI ARG_2_HI OVERFLOW) (+ RES_HI (* 340282366920938463463374607431768211456 (shift OVERFLOW -1)))))) (if (- INST 1) 0 (begin (- (+ RES_LO ARG_2_LO) (+ ARG_1_LO (* 340282366920938463463374607431768211456 OVERFLOW))) (- (+ RES_HI ARG_2_HI OVERFLOW) (+ ARG_1_HI (* 340282366920938463463374607431768211456 (shift OVERFLOW -1))))))))))

(defconstraint stamp-constancies () (begin (if (- (shift STAMP 1) STAMP) (- (shift ARG_1_HI 1) ARG_1_HI)) (if (- (shift STAMP 1) STAMP) (- (shift ARG_1_LO 1) ARG_1_LO)) (if (- (shift STAMP 1) STAMP) (- (shift ARG_2_HI 1) ARG_2_HI)) (if (- (shift STAMP 1) STAMP) (- (shift ARG_2_LO 1) ARG_2_LO)) (if (- (shift STAMP 1) STAMP) (- (shift RES_HI 1) RES_HI)) (if (- (shift STAMP 1) STAMP) (- (shift RES_LO 1) RES_LO)) (if (- (shift STAMP 1) STAMP) (- (shift INST 1) INST)) (if (- (shift STAMP 1) STAMP) (- (shift CT_MAX 1) CT_MAX))))

(defconstraint heartbeat () (begin (if STAMP (begin INST)) (* (- (shift STAMP 1) STAMP) (- (shift STAMP 1) (+ STAMP 1))) (if (- (shift STAMP 1) STAMP) 0 (shift CT 1)) (if STAMP 0 (begin (* (- INST 1) (- INST 3)) (if (- 1 (~ (- CT CT_MAX))) (- (shift CT 1) (+ CT 1)) (- (shift STAMP 1) (+ STAMP 1))) (- (~ (* (- CT 16) CT_MAX)) 1)))))

(defconstraint binary-and-byte-decompositions () (begin (if CT (- ACC_1 BYTE_1) (- ACC_1 (+ (* 256 (shift ACC_1 -1)) BYTE_1))) (if CT (- ACC_2 BYTE_2) (- ACC_2 (+ (* 256 (shift ACC_2 -1)) BYTE_2)))))

(defconstraint last-row (:domain {-1}) (- CT CT_MAX))

(defconstraint first-row (:domain {0}) STAMP)
