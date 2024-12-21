(defconst
  EMPTY_KECCAK_HI                           0xc5d2460186f7233c927e7db2dcc703c0
  EMPTY_KECCAK_LO                           0xe500b653ca82273b7bfad8045d85a470
  EVM_INST_PUSH1                            0x60
  EVM_INST_INVALID                          0xFE
  ;;
  LLARGE                                    16
  LLARGEMO                                  (- LLARGE 1)
  LLARGEPO                                  (+ LLARGE 1)
  WORD_SIZE                                 32
  WORD_SIZE_MO                              (- WORD_SIZE 1))

(module rom)

(defcolumns
  (CODE_FRAGMENT_INDEX :i32)
  (CODE_FRAGMENT_INDEX_INFTY :i32)
  (CODE_SIZE :i32 :display :dec)
  (CODESIZE_REACHED :binary@prove)
  (PROGRAM_COUNTER :i32)
  (LIMB :i128)
  (nBYTES :byte)
  (nBYTES_ACC :byte)
  (INDEX :i32)
  (COUNTER :byte)
  (COUNTER_MAX :byte)
  (PADDED_BYTECODE_BYTE :byte@prove)
  (ACC :i128)
  (IS_PUSH :binary)
  (PUSH_PARAMETER :byte)
  (COUNTER_PUSH :byte)
  (IS_PUSH_DATA :binary@prove)
  (PUSH_VALUE_HI :i128)
  (PUSH_VALUE_LO :i128)
  (PUSH_VALUE_ACC :i128)
  (PUSH_FUNNEL_BIT :binary@prove)
  (OPCODE :byte :display :opcode)
  (IS_JUMPDEST :binary))

(defalias
  PC   PROGRAM_COUNTER
  CFI  CODE_FRAGMENT_INDEX
  CT   COUNTER
  PBCB PADDED_BYTECODE_BYTE)

(module rom)

(defpurefun (if-not-eq A B then)
  (if (neq A B)
      then))

;; Constancies
(defun (cfi-constant X)
  (if-not-eq CFI
             (+ (prev CFI) 1)
             (remained-constant! X)))

(defun (cfi-incrementing X)
  (if-not-eq CFI
             (+ (prev CFI) 1)
             (or! (remained-constant! X) (did-inc! X 1))))

(defpurefun (counter-constant X ct ctmax)
  (if-not-eq ct ctmax (will-remain-constant! X)))

(defun (push-constant X)
  (if-not-zero COUNTER_PUSH
               (remained-constant! X)))

(defun (push-incrementing X)
  (if-not-zero COUNTER_PUSH
               (or! (remained-constant! X) (did-inc! X 1))))

(defconstraint cfi-constancies ()
  (cfi-constant CODE_SIZE))

(defconstraint cfi-incrementings ()
  (begin (cfi-incrementing CODESIZE_REACHED)
         (debug (cfi-incrementing PC))))

(defconstraint ct-constancies ()
  (begin (counter-constant LIMB CT COUNTER_MAX)
         (counter-constant nBYTES CT COUNTER_MAX)
         (counter-constant COUNTER_MAX CT COUNTER_MAX)))

(defconstraint push-constancies ()
  (begin (push-constant PUSH_PARAMETER)
         (push-constant PUSH_VALUE_HI)
         (push-constant PUSH_VALUE_LO)))

;; Heartbeat
(defconstraint initialization (:domain {0})
  (vanishes! CODE_FRAGMENT_INDEX))

(defconstraint cfi-evolving-possibility ()
  (or! (will-remain-constant! CFI) (will-inc! CFI 1)))

(defconstraint no-cfi-nothing ()
  (if-zero CFI
           (begin (vanishes! CT)
                  (vanishes! COUNTER_MAX)
                  (vanishes! PBCB)
                  (debug (vanishes! IS_PUSH))
                  (debug (vanishes! IS_PUSH_DATA))
                  (debug (vanishes! COUNTER_PUSH))
                  (debug (vanishes! PUSH_PARAMETER))
                  (debug (vanishes! PROGRAM_COUNTER)))
           (begin (debug (or! (eq! COUNTER_MAX LLARGEMO) (eq! COUNTER_MAX WORD_SIZE_MO)))
                  (if-eq COUNTER_MAX LLARGEMO (will-remain-constant! CFI))
                  (if-not-eq COUNTER COUNTER_MAX (will-remain-constant! CFI))
                  (if-eq CT WORD_SIZE_MO (will-inc! CFI 1)))))

(defconstraint counter-evolution ()
  (if-eq-else CT COUNTER_MAX
              (vanishes! (next CT))
              (will-inc! CT 1)))

(defconstraint finalisation (:domain {-1})
  (if-not-zero CFI
               (begin (eq! CT COUNTER_MAX)
                      (eq! COUNTER_MAX WORD_SIZE_MO)
                      (eq! CFI CODE_FRAGMENT_INDEX_INFTY))))

(defconstraint cfi-infty ()
  (if-zero CFI
           (vanishes! CODE_FRAGMENT_INDEX_INFTY)
           (will-remain-constant! CODE_FRAGMENT_INDEX_INFTY)))

(defconstraint limb-accumulator ()
  (begin (if-zero CT
                  (eq! ACC PBCB)
                  (eq! ACC
                       (+ (* 256 (prev ACC))
                          PBCB)))
         (if-eq CT COUNTER_MAX (eq! ACC LIMB))))

;; CODESIZE_REACHED Constraints
(defconstraint codesizereached-trigger ()
  (if-eq PC (- CODE_SIZE 1)
         (eq! (+ CODESIZE_REACHED (next CODESIZE_REACHED))
              1)))

(defconstraint csr-impose-ctmax (:guard CFI)
  (if-zero CT
           (if-zero CODESIZE_REACHED
                    (eq! COUNTER_MAX LLARGEMO)
                    (eq! COUNTER_MAX WORD_SIZE_MO))))

;; nBytes constraints
(defconstraint nbytes-acc (:guard CFI)
  (if-zero CT
           (if-zero CODESIZE_REACHED
                    (eq! nBYTES_ACC 1)
                    (vanishes! nBYTES))
           (if-zero CODESIZE_REACHED
                    (did-inc! nBYTES_ACC 1)
                    (remained-constant! nBYTES_ACC))))

(defconstraint nbytes-collusion ()
  (if-eq CT COUNTER_MAX (eq! nBYTES nBYTES_ACC)))

;; INDEX constraints
(defconstraint no-cfi-no-index ()
  (if-zero CFI
           (vanishes! INDEX)))

(defconstraint new-cfi-reboot-index ()
  (if-not-zero (- CFI (prev CFI))
               (vanishes! INDEX)))

(defconstraint new-ct-increment-index ()
  (if-not-zero (any! CFI
                     (did-inc! CFI 1)
                     (- 1 (~ CT)))
               (did-inc! INDEX 1)))

(defconstraint index-inc-in-middle-padding ()
  (if-eq CT LLARGE (did-inc! INDEX 1)))

(defconstraint index-quasi-ct-cst ()
  (if-not-zero (* CT (- CT LLARGE))
               (remained-constant! INDEX)))

;; PC constraints
(defconstraint pc-incrementing (:guard CFI)
  (if-not-eq (next CFI) (+ CFI 1) (will-inc! PC 1)))

(defconstraint pc-reboot ()
  (if-not-eq (next CFI)
             CFI
             (vanishes! (next PC))))

;; end of CFI (padding rows)
(defconstraint end-code-no-opcode ()
  (if-eq CODESIZE_REACHED 1 (vanishes! PBCB)))

;; Constraints Related to PUSHX instructions
(defconstraint not-a-push-data ()
  (if-zero IS_PUSH_DATA
           (begin (vanishes! COUNTER_PUSH)
                  (eq! OPCODE PBCB))))

(defconstraint ispush-ispushdata-exclusivity ()
  (vanishes! (* IS_PUSH IS_PUSH_DATA)))

(defconstraint ispush-implies-next-pushdata ()
  (if-not-zero IS_PUSH (eq! (next IS_PUSH_DATA) 1)))

(defconstraint ispush-constraint ()
  (if-not-zero IS_PUSH
               (begin (eq! PUSH_PARAMETER
                           (- OPCODE (- EVM_INST_PUSH1 1)))
                      (vanishes! PUSH_VALUE_ACC)
                      (vanishes! (+ PUSH_FUNNEL_BIT (next PUSH_FUNNEL_BIT))))))

(defconstraint ispushdata-constraint ()
  (if-not-zero IS_PUSH_DATA
               (begin (eq! (+ (prev IS_PUSH) (prev IS_PUSH_DATA))
                           1)
                      (eq! OPCODE EVM_INST_INVALID)
                      (did-inc! COUNTER_PUSH 1)
                      (if-zero (- (+ COUNTER_PUSH LLARGE) PUSH_PARAMETER)
                               (begin (will-inc! PUSH_FUNNEL_BIT 1)
                                      (eq! PUSH_VALUE_HI PUSH_VALUE_ACC))
                               (if-eq (next IS_PUSH_DATA) 1 (will-remain-constant! PUSH_FUNNEL_BIT)))
                      (if-zero (- (prev PUSH_FUNNEL_BIT) PUSH_FUNNEL_BIT)
                               (eq! PUSH_VALUE_ACC
                                    (+ (* 256 (prev PUSH_VALUE_ACC))
                                       PBCB))
                               (eq! PUSH_VALUE_ACC PBCB))
                      (if-eq COUNTER_PUSH PUSH_PARAMETER
                             (begin (if-zero PUSH_FUNNEL_BIT
                                             (vanishes! PUSH_VALUE_HI))
                                    (eq! PUSH_VALUE_ACC PUSH_VALUE_LO)
                                    (vanishes! (next IS_PUSH_DATA)))))))


(module romlex)

(defcolumns
  (CODE_FRAGMENT_INDEX :i32)
  (CODE_FRAGMENT_INDEX_INFTY :i32)
  (CODE_SIZE :i32)
  (ADDRESS_HI :i32)
  (ADDRESS_LO :i128)
  (DEPLOYMENT_NUMBER :i16)
  (DEPLOYMENT_STATUS :binary@prove)
  (CODE_HASH_HI :i128 :display :hex)
  (CODE_HASH_LO :i128 :display :hex)
  (COMMIT_TO_STATE :binary@prove)
  (READ_FROM_STATE :binary@prove))

(defalias
  CFI CODE_FRAGMENT_INDEX)


(module romlex)

(defconstraint initialization (:domain {0})
  (vanishes! CODE_FRAGMENT_INDEX))

(defconstraint cfi-evolution ()
  (or! (will-inc! CFI 1) (will-remain-constant! CFI)))

(defconstraint finalisation (:domain {-1})
  (if-not-zero CFI
               (eq! CFI CODE_FRAGMENT_INDEX_INFTY)))

(defconstraint cfi-rules ()
  (if-zero CFI
           (vanishes! CODE_FRAGMENT_INDEX_INFTY)
           (begin (will-inc! CFI 1)
                  (will-remain-constant! CODE_FRAGMENT_INDEX_INFTY))))

(defconstraint keccak-of-initcode (:guard DEPLOYMENT_STATUS)
  (begin (eq! CODE_HASH_HI EMPTY_KECCAK_HI)
         (eq! CODE_HASH_LO EMPTY_KECCAK_LO)))

;; TODO add lexicographic ordering


(deflookup
  romlex-into-rom
  ;; target columns
  (
    rom.CODE_FRAGMENT_INDEX
    rom.CODE_FRAGMENT_INDEX_INFTY
    rom.CODE_SIZE
  )
  ;; source columns
  (
    romlex.CODE_FRAGMENT_INDEX
    romlex.CODE_FRAGMENT_INDEX_INFTY
    romlex.CODE_SIZE
  ))
