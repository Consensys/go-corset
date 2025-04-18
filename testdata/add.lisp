(module add)

(defcolumns
  (STAMP :i32)
  (CT_MAX :byte)
  (CT :byte)
  (INST :byte :display :opcode)
  (ARG_1_HI :i128)
  (ARG_1_LO :i128)
  (ARG_2_HI :i128)
  (ARG_2_LO :i128)
  (RES_HI :i128)
  (RES_LO :i128)
  (BYTE_1 :byte@prove)
  (BYTE_2 :byte@prove)
  (ACC_1 :i128)
  (ACC_2 :i128)
  (OVERFLOW :binary@prove))

(defconst
  EVM_INST_STOP                          0x00
  EVM_INST_ADD                           0x01
  EVM_INST_MUL                           0x02
  EVM_INST_SUB                           0x03
  LLARGEMO 15
  LLARGE   16
  THETA 340282366920938463463374607431768211456) ;; note that 340282366920938463463374607431768211456 = 256^16

(defconstraint stamp-constancies ()
  (begin (stamp-constancy STAMP ARG_1_HI)
         (stamp-constancy STAMP ARG_1_LO)
         (stamp-constancy STAMP ARG_2_HI)
         (stamp-constancy STAMP ARG_2_LO)
         (stamp-constancy STAMP RES_HI)
         (stamp-constancy STAMP RES_LO)
         (stamp-constancy STAMP INST)
         (stamp-constancy STAMP CT_MAX)))

;;;;;;;;;;;;;;;;;;;;;;;;;
;;                     ;;
;;    1.3 heartbeat    ;;
;;                     ;;
;;;;;;;;;;;;;;;;;;;;;;;;;
(defconstraint first-row (:domain {0})
  (vanishes! STAMP))

(defconstraint heartbeat ()
  (begin (if-zero STAMP
                  (begin (vanishes! INST)
                         ;; (debug (vanishes! CT))
                         ;; (debug (vanishes! CT_MAX))
                         ))
         (or! (will-remain-constant! STAMP) (will-inc! STAMP 1))
         (if-not (will-remain-constant! STAMP)
                 (vanishes! (next CT)))
         (if-not-zero STAMP
                      (begin (or! (eq! INST EVM_INST_ADD) (eq! INST EVM_INST_SUB))
                             (if-eq-else CT CT_MAX
                                 (will-inc! STAMP 1)
                                 (will-inc! CT 1))
                             (eq! (~ (* (- CT LLARGE) CT_MAX))
                                  1)))))

(defconstraint last-row (:domain {-1})
  (eq! CT CT_MAX))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;                                                   ;;
;;    1.4 binary, bytehood and byte decompositions   ;;
;;                                                   ;;
;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
(defconstraint binary-and-byte-decompositions ()
  (begin (byte-decomposition CT ACC_1 BYTE_1)
         (byte-decomposition CT ACC_2 BYTE_2)))

;; TODO: bytehood constraints
;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;                       ;;
;;    1.5 constraints    ;;
;;                       ;;
;;;;;;;;;;;;;;;;;;;;;;;;;;;
(defconstraint adder-constraints (:guard STAMP)
  (if-eq CT CT_MAX
         (begin (eq! RES_HI ACC_1)
                (eq! RES_LO ACC_2)
                (if-not-zero (- INST EVM_INST_SUB)
                             (begin (eq! (+ ARG_1_LO ARG_2_LO)
                                         (+ RES_LO (* THETA OVERFLOW)))
                                    (eq! (+ ARG_1_HI ARG_2_HI OVERFLOW)
                                         (+ RES_HI
                                            (* THETA (prev OVERFLOW))))))
                (if-not-zero (- INST EVM_INST_ADD)
                             (begin (eq! (+ RES_LO ARG_2_LO)
                                         (+ ARG_1_LO (* THETA OVERFLOW)))
                                    (eq! (+ RES_HI ARG_2_HI OVERFLOW)
                                         (+ ARG_1_HI
                                            (* THETA (prev OVERFLOW)))))))))
