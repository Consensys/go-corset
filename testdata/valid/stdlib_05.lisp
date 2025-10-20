(defcolumns (A :byte) (B :binary@prove) (T :binary@prove))

(defconstraint old ()
  ;; if A==1 && B == 0
  (if-zero (+ (~ (- A 1)) B)
           ;; then T == 1
           (eq! 1 T)
           ;; else T == 0
           (== 0 T)))

(defconstraint new ()
  ;; if A==1 && B == 0
  (if (and! (eq! A 1) (eq! B 0))
           ;; then T == 1
           (eq! T 1)
           ;; else T == 0
           (== 0 T)))
