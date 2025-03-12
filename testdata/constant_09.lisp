(defpurefun ((eq! :𝔽@loob) x y) (- x y))
(defpurefun (if-eq x val then) (if (eq! x val) then))
;;
(defcolumns
  (CT :byte)
  (IS_SLT :binary@prove)
  (BITS :binary@prove)
  (NEG_1 :binary@prove)
  (NEG_2 :binary@prove)
  (BYTE_1 :byte@prove)
  (BYTE_3 :byte@prove)
  )

;; opcode values
(defconst
  LLARGE                                    16
  LLARGEMO                                  (- LLARGE 1))

(defconstraint bits-and-negs (:guard IS_SLT)
  (if-eq CT LLARGEMO
	 (eq! NEG_2 (shift BITS (- 0 7)))))
