(column NIBBLE :u4)
(column BIT_0 :u1)
(column BIT_1 :u1)
(column BIT_2 :u1)
(column BIT_3 :u1)

;; NIBBLE = 8*BIT_3 + 4*BIT_2 + 2*BIT_1 + BIT_0
(vanish decomp (- NIBBLE (+ BIT_0 (* 2 BIT_1) (* 4 BIT_2) (* 8 BIT_3))))
