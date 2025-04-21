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
  (== STAMP 0))

(defconstraint heartbeat ()
  (begin
   (if (== STAMP 0)
       (== INST 0))
   ;; Stamp either constant is increases by 1
   (∨ (will-remain-constant! STAMP) (will-inc! STAMP 1))
   ;; When stamp increases, counter is reset
   (if (¬ (will-remain-constant! STAMP))
       (== (next CT) 0))
   ;;
   (if (!= STAMP 0)
       (begin
        ;; outside of padding, instruction either ADD or SUB
        (∨ (eq! INST EVM_INST_ADD) (eq! INST EVM_INST_SUB))
        ;;
        (if (== CT CT_MAX)
            ;; After last row of frame, stamp increases
            (will-inc! STAMP 1)
            ;; On rows within frame, counter increases
            (will-inc! CT 1))
        ;; (CT < LLARGE) ∧ (CT_MAX > 0)
        (∧ (!= CT LLARGE) (!= CT_MAX 0))))))

(defconstraint last-row (:domain {-1})
  (== CT CT_MAX))

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
  (if (== CT CT_MAX)
      (begin
       ;;
       (== RES_HI ACC_1)
       (== RES_LO ACC_2)
       ;;
       (if (!= INST EVM_INST_SUB)
           (begin (== (+ ARG_1_LO ARG_2_LO)
                      (+ RES_LO (* THETA OVERFLOW)))
                  (== (+ ARG_1_HI ARG_2_HI OVERFLOW)
                      (+ RES_HI
                          (* THETA (prev OVERFLOW))))))
       ;;
       (if (!= INST EVM_INST_ADD)
           (begin (== (+ RES_LO ARG_2_LO)
                      (+ ARG_1_LO (* THETA OVERFLOW)))
                  (== (+ RES_HI ARG_2_HI OVERFLOW)
                      (+ ARG_1_HI
                         (* THETA (prev OVERFLOW)))))))))
