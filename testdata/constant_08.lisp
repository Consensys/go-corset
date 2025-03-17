(defpurefun (eq! x y) (== x y))

(defcolumns
  (C :byte)
  (L :binary)
  (B :binary)
  (N :binary))

;; opcode values
(defconst
  LLARGE                                    16
  LLARGEMO                                  (- LLARGE 1))

(defconstraint bits-and-negs (:guard L)
  (if (eq! C LLARGEMO)
      (eq! N
	   (shift B (- 0 LLARGEMO)))))
