(defcolumns
  (add:ACC_2 :i128)
  (add:RES_LO :i128)
  (add:ARG_1_LO :i128)
  (add:OVERFLOW :binary@prove)
  (add:RES_HI :i128)
  (add:INST :i8)
  (add:BYTE_1 :byte@prove)
  (add:BYTE_2 :byte@prove)
  (add:ACC_1 :i128)
  (add:STAMP :i32)
  (add:ARG_1_HI :i128)
  (add:ARG_2_LO :i128)
  (add:ARG_2_HI :i128)
  (add:CT_MAX :i8)
  (add:CT :i8))

(defconstraint add:adder-constraints () (if add:STAMP 0 (if (- add:CT add:CT_MAX) (begin (- add:RES_HI add:ACC_1) (- add:RES_LO add:ACC_2) (if (- add:INST 3) 0 (begin (- (+ add:ARG_1_LO add:ARG_2_LO) (+ add:RES_LO (* 340282366920938463463374607431768211456 add:OVERFLOW))) (- (+ add:ARG_1_HI add:ARG_2_HI add:OVERFLOW) (+ add:RES_HI (* 340282366920938463463374607431768211456 (shift add:OVERFLOW -1)))))) (if (- add:INST 1) 0 (begin (- (+ add:RES_LO add:ARG_2_LO) (+ add:ARG_1_LO (* 340282366920938463463374607431768211456 add:OVERFLOW))) (- (+ add:RES_HI add:ARG_2_HI add:OVERFLOW) (+ add:ARG_1_HI (* 340282366920938463463374607431768211456 (shift add:OVERFLOW -1))))))))))

(defconstraint add:stamp-constancies () (begin (if (- (shift add:STAMP 1) add:STAMP) (- (shift add:ARG_1_HI 1) add:ARG_1_HI)) (if (- (shift add:STAMP 1) add:STAMP) (- (shift add:ARG_1_LO 1) add:ARG_1_LO)) (if (- (shift add:STAMP 1) add:STAMP) (- (shift add:ARG_2_HI 1) add:ARG_2_HI)) (if (- (shift add:STAMP 1) add:STAMP) (- (shift add:ARG_2_LO 1) add:ARG_2_LO)) (if (- (shift add:STAMP 1) add:STAMP) (- (shift add:RES_HI 1) add:RES_HI)) (if (- (shift add:STAMP 1) add:STAMP) (- (shift add:RES_LO 1) add:RES_LO)) (if (- (shift add:STAMP 1) add:STAMP) (- (shift add:INST 1) add:INST)) (if (- (shift add:STAMP 1) add:STAMP) (- (shift add:CT_MAX 1) add:CT_MAX))))

(defconstraint add:heartbeat () (begin (if add:STAMP (begin add:INST)) (* (- (shift add:STAMP 1) add:STAMP) (- (shift add:STAMP 1) (+ add:STAMP 1))) (if (- (shift add:STAMP 1) add:STAMP) 0 (shift add:CT 1)) (if add:STAMP 0 (begin (* (- add:INST 1) (- add:INST 3)) (if (- 1 (~ (- add:CT add:CT_MAX))) (- (shift add:CT 1) (+ add:CT 1)) (- (shift add:STAMP 1) (+ add:STAMP 1))) (- (~ (* (- add:CT 16) add:CT_MAX)) 1)))))

(defconstraint add:binary-and-byte-decompositions () (begin (if add:CT (- add:ACC_1 add:BYTE_1) (- add:ACC_1 (+ (* 256 (shift add:ACC_1 -1)) add:BYTE_1))) (if add:CT (- add:ACC_2 add:BYTE_2) (- add:ACC_2 (+ (* 256 (shift add:ACC_2 -1)) add:BYTE_2)))))

(defconstraint add:last-row (:domain {-1}) (- add:CT add:CT_MAX))

(defconstraint add:first-row (:domain {0}) add:STAMP)
